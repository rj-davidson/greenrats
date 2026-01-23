"use client";

import { Badge } from "@/components/shadcn/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useAdminUsers } from "@/features/admin/queries";

export function UsersTable() {
  const { data, isLoading, error } = useAdminUsers();

  if (isLoading) {
    return <div className="py-4 text-muted-foreground">Loading users...</div>;
  }

  if (error) {
    return <div className="py-4 text-destructive">Failed to load users</div>;
  }

  if (!data || data.users.length === 0) {
    return <div className="py-4 text-muted-foreground">No users found</div>;
  }

  return (
    <div className="space-y-4">
      <div className="text-sm text-muted-foreground">{data.total} total users</div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Email</TableHead>
            <TableHead>Display Name</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.users.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="font-medium">{user.email}</TableCell>
              <TableCell>{user.display_name ?? "-"}</TableCell>
              <TableCell>
                {user.is_admin ? (
                  <Badge variant="default">Admin</Badge>
                ) : (
                  <Badge variant="secondary">User</Badge>
                )}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {new Date(user.created_at).toLocaleDateString()}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
