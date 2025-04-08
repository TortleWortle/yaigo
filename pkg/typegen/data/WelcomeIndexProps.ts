export type WelcomeIndexProps = {
	basic: Typegen_testBasicStruct;
	basicButOptional?: Typegen_testBasicStruct;
	basicStructSlice: Typegen_testBasicStruct[];
	intProp: number;
	nilField: null;
	optInt?: number;
	optString?: string;
	optStringSlice?: string[];
	otherPkgStructMap: {
		[key: string]: DbGroup;
	};
	otherPkgStructSlice: DbUser[];
	stringProp: string;
	stringSlice: string[];
}
